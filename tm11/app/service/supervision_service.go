package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm11/app/model"
	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/helper"
)

type SupervisionService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisionService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisionService {
	return &SupervisionService{supervisors: supervisors, tx: tx}
}

// ---------- GET /supervisors ----------
func (s *SupervisionService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisionService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	sv, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", sv)
}

// ---------- POST /supervisors/:id/assign ----------
// Dijaga permission supervision:manage.
//
// Dua jaminan, dua teknik:
//   - kuota pembimbing  -> UPDATE bersyarat atomik (IncrementSupervised)
//   - satu mahasiswa satu pembimbing -> kunci row students + PRIMARY KEY
//     pada supervisions.student_id sebagai jaring pengaman database.
func (s *SupervisionService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.AssignRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var assigned model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci mahasiswa: request lain untuk mahasiswa yang SAMA antre.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}
		// 2. Ambil satu slot secara atomik.
		var err error
		assigned, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}
		// 3. Catat. Bila mahasiswa sudah punya pembimbing, error di sini
		//    membatalkan juga penambahan slot pada langkah 2.
		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Created(c, "pembimbing berhasil ditetapkan", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
// Dijaga permission supervision:manage.
//
// Atomik: bila pembimbing baru penuh, seluruh transaction dibatalkan dan
// mahasiswa TETAP dengan pembimbing lamanya. Bebas deadlock: kedua
// pembimbing dikunci dalam urutan id menaik.
func (s *SupervisionService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var joined model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci mahasiswa DULU. Setelah ini, pembimbing "saat ini"
		//    milik mahasiswa tidak mungkin berubah oleh request lain.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}
		fromID, err := repo.CurrentSupervisorID(ctx, req.StudentID)
		if err != nil {
			return err
		}
		if fromID == req.ToSupervisorID {
			return errSameSupervisor
		}

		// 2. Kunci KEDUA pembimbing dalam urutan id menaik.
		if err := repo.LockSupervisors(ctx, fromID, req.ToSupervisorID); err != nil {
			return err
		}

		// 3. Keluar dari yang lama, masuk ke yang baru.
		if err := repo.DeleteSupervision(ctx, req.StudentID); err != nil {
			return err
		}
		if _, err := repo.DecrementSupervised(ctx, fromID); err != nil {
			return err
		}
		joined, err = repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}
		return repo.InsertSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "berhasil ganti pembimbing", joined)
}

// ---------- PATCH /supervisors/:id ----------
// Optimistic locking, dijaga permission supervision:manage.
func (s *SupervisionService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

var errSameSupervisor = errors.New("pembimbing baru sama dengan pembimbing lama")

func translateSupervisionError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("mahasiswa atau pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki pembimbing")
	case errors.Is(err, errSameSupervisor):
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan pembimbing saat ini",
		})
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrSupervisorQuotaBelowCnt):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}
