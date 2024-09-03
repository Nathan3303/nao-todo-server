const mongoose = require("mongoose");

const SessionSchema = new mongoose.Schema({
    userId: { type: String, required: true },
    token: { type: String, required: true },
    createdAt: { type: Date, default: Date.now },
    expiresAt: { type: Date, required: true },
    // isActive: { type: Boolean, default: true },
    browser: { type: String, required: true },
    // region: { type: String, required: true },
    ip: { type: String },
    device: { type: String },
    os: { type: String },
});

const Session = mongoose.model("Session", SessionSchema);

module.exports = Session;
