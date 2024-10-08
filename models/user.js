const mongoose = require("mongoose");

const UserSchema = new mongoose.Schema({
    username: { type: String, required: true, unique: true },
    password: { type: String, required: true },
    email: { type: String, required: true, unique: true },
    nickName: { type: String },
    firstName: { type: String },
    lastName: { type: String },
    role: { type: String, required: true, default: "user" },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now },
    isDeleted: { type: Boolean, default: false },
    lastSignIn: { type: Date },
    lastSignOut: { type: Date },
});

const User = mongoose.model("User", UserSchema);

module.exports = User;
