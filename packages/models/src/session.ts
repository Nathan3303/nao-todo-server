import { Schema, model } from 'mongoose';

const SessionSchema = new Schema({
    userId: { type: String, required: true },
    token: { type: String, required: true },
    createdAt: { type: Date, default: Date.now },
    expiresAt: { type: Date, required: true },
    browser: { type: String, required: true },
    ip: { type: String },
    device: { type: String },
    os: { type: String }
});

export default model('Session', SessionSchema);
