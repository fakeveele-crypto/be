// Eksplorasi modul: bandingkan totalDocsExamined DENGAN dan TANPA index.
//   mongosh --file tm10/mongo/explain_tanpa_index.js
// Memakai collection tiruan (latihan_mongo.ulasan_demo, 5000 dokumen) agar
// data aplikasi tidak disentuh dan perbedaannya terlihat jelas.
db = db.getSiblingDB("latihan_mongo");
db.ulasan_demo.drop();

const docs = [];
for (let i = 0; i < 5000; i++) {
  docs.push({ course_id: (i % 5) + 1, user_id: i, rating: (i % 5) + 1, tags: i % 3 === 0 ? ["praktikal"] : ["seru"] });
}
db.ulasan_demo.insertMany(docs);
db.ulasan_demo.createIndex({ course_id: 1, _id: -1 }, { name: "course_newest" });

function periksa(judul) {
  const ex = db.ulasan_demo.find({ course_id: 1 }).explain("executionStats");
  const s = ex.executionStats;
  print(judul);
  print("  dokumen diperiksa   : " + s.totalDocsExamined);
  print("  dokumen dikembalikan: " + s.nReturned + "\n");
}

periksa("DENGAN index course_newest");
db.ulasan_demo.dropIndex("course_newest");
periksa("TANPA index (collection scan)");
db.ulasan_demo.drop();
