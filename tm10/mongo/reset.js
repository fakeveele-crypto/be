// Mengosongkan data uji Modul 10 di database aplikasi. Index TIDAK dihapus.
//   mongosh --file tm10/mongo/reset.js
db = db.getSiblingDB("praktikum_backend");
print("course_reviews dihapus     : " + db.course_reviews.deleteMany({}).deletedCount);
print("student_portfolios dihapus : " + db.student_portfolios.deleteMany({}).deletedCount);
