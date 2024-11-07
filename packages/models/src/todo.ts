import { Schema, model } from 'mongoose';

const ObjectId = Schema.Types.ObjectId;

const todoSchema = new Schema({
    userId: { type: ObjectId, ref: "User" },
    projectId: { type: ObjectId, ref: "Project" },
    name: { type: String, required: true },
    description: { type: String, default: "" },
    state: { type: String, default: "todo" },
    priority: { type: String, default: "low" },
    tags: [{ type: String }],
    dueDate: {
        startAt: { type: Date, default: Date.now },
        endAt: { type: Date, default: null },
    },
    isDone: { type: Boolean, default: false },
    isDeleted: { type: Boolean, default: false },
    isArchived: { type: Boolean, default: false },
    isPinned: { type: Boolean, default: false },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now },
});

export default model("Todo", todoSchema);
