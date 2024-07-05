const mongoose = require("mongoose");

const ObjectId = mongoose.Schema.Types.ObjectId;

const todoSchema = new mongoose.Schema({
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

const Todo = mongoose.model("Todo", todoSchema);

module.exports = Todo;
