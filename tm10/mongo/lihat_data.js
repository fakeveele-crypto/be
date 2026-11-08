// Menampilkan satu dokumen ulasan dan satu dokumen portofolio dari database
// aplikasi, untuk dibandingkan dengan response API (dan untuk tangkapan layar
// bila Compass tidak tersedia). Jalankan:
//   mongosh --file tm10/mongo/lihat_data.js
db = db.getSiblingDB("praktikum_backend");

print("== course_reviews: jumlah dokumen " + db.course_reviews.countDocuments() + " ==");
printjson(db.course_reviews.findOne({}, {}, { sort: { _id: -1 } }));

print("== student_portfolios: jumlah dokumen " + db.student_portfolios.countDocuments() + " ==");
printjson(db.student_portfolios.findOne({}, {}, { sort: { _id: -1 } }));

print("== index course_reviews ==");
printjson(db.course_reviews.getIndexes().map(i => ({ nama: i.name, kunci: i.key, unik: !!i.unique })));
print("== index student_portfolios ==");
printjson(db.student_portfolios.getIndexes().map(i => ({ nama: i.name, kunci: i.key, unik: !!i.unique })));
