import { Schema, model } from 'mongoose';

const ObjectId = Schema.Types.ObjectId;

const tagSchema = new Schema({
    userId: { type: ObjectId, ref: 'User' },
    name: { type: String, required: true },
    color: { type: String, required: true },
    isDeleted: { type: Boolean, default: false },
    deletedAt: { type: Date },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});

export default model('Tag', tagSchema);
