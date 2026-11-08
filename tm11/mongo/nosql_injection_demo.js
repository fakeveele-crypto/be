// Demo NoSQL injection (bagian A.8 modul). Dijalankan dengan:
//   mongosh --file tm10/mongo/nosql_injection_demo.js
db = db.getSiblingDB("latihan_mongo");
db.akun_demo.drop();
db.akun_demo.insertOne({ username: "admin", password: "rahasia-sekali" });

// Seolah-olah inilah isi body JSON yang dikirim penyerang
const body = JSON.parse('{"username": "admin", "password": {"$ne": ""}}');

print("== query yang menerima body apa adanya ==");
printjson(db.akun_demo.findOne({ username: body.username, password: body.password }));

print("== query yang memaksa password bertipe string ==");
const password = typeof body.password === "string" ? body.password : null;
printjson(db.akun_demo.findOne({ username: body.username, password: password }));
