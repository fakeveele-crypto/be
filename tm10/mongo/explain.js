const COURSE_ID = 1;      // kelas yang punya ulasan
const TAG = "uji";        // salah satu tag yang ada pada ulasan kelas itu
const STUDENT_ID = 20052;     // students.id yang sudah punya portofolio

db = db.getSiblingDB("pbl");

function cariIndex(node, hasil = []) {
  if (node && typeof node === "object") {
    if (/IXSCAN$/.test(node.stage || "") && node.indexName) hasil.push(node.indexName);
    for (const k of Object.keys(node)) cariIndex(node[k], hasil);
  }
  return hasil;
}
function laporkan(judul, ex) {
  const s = ex.executionStats;
  const idx = [...new Set(cariIndex(ex.queryPlanner.winningPlan))];
  print(judul);
  print("  index dipakai       : " + (idx.length ? idx.join(", ") : "TIDAK ADA (COLLSCAN)"));
  print("  dokumen diperiksa   : " + s.totalDocsExamined);
  print("  dokumen dikembalikan: " + s.nReturned);
  print("  kunci diperiksa     : " + s.totalKeysExamined + "\n");
}

laporkan("daftar ulasan satu kelas (terbaru dulu)",
  db.course_reviews.find({ course_id: COURSE_ID }).sort({ _id: -1 }).explain("executionStats"));
laporkan("filter tag pada satu kelas",
  db.course_reviews.find({ course_id: COURSE_ID, tags: TAG }).explain("executionStats"));
laporkan("ulasan milik satu mahasiswa di satu kelas (index unik)",
  db.course_reviews.find({ course_id: COURSE_ID, user_id: 1 }).explain("executionStats"));
laporkan("portofolio berdasarkan student_id (index unik student_unique)",
  db.student_portfolios.find({ student_id: STUDENT_ID }).explain("executionStats"));
laporkan("kondisi batas proyek (filter student_id + projects.19 tidak ada)",
  db.student_portfolios.find({ student_id: STUDENT_ID, "projects.19": { $exists: false } }).explain("executionStats"));
