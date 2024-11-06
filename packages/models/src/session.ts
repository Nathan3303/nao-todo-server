import mongoose from "mongoose";

const SessionSchema = new mongoose.Schema({
    userId: { type: String, required: true },
    token: { type: String, required: true },
    createdAt: { type: Date, default: Date.now },
    expiresAt: { type: Date, required: true },
    browser: { type: String, required: true },
    ip: { type: String },
    device: { type: String },
    os: { type: String },
});

export default mongoose.model("Session", SessionSchema);
