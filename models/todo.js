const mongoose = require("mongoose");

const ObjectId = mongoose.Schema.Types.ObjectId;

const todoSchema = new mongoose.Schema({
    projectId: { type: ObjectId, ref: "Project" },
    name: { type: String, required: true },
    description: { type: String },
    state: { type: String, default: "todo" },
    priority: { type: Number, default: 0 },
    tags: [{ type: String }],
    dueData: { type: Date },
    isDone: { type: Boolean, default: false },
    isDeleted: { type: Boolean, default: false },
    isArchived: { type: Boolean, default: false },
    createAt: { type: Date, default: Date.now },
    updateAt: { type: Date, default: Date.now },
});

const Todo = mongoose.model("Todo", todoSchema);

module.exports = Todo;
