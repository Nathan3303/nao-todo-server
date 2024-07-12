const mongoose = require("mongoose");

const ObjectId = mongoose.Schema.Types.ObjectId;

const eventSchema = new mongoose.Schema({
    todoId: { type: ObjectId, ref: "Todo" },
    title: { type: String, required: true },
    isDone: { type: Boolean, default: false },
    isTopped: { type: Boolean, default: false },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now },
});

const Event = mongoose.model("Event", eventSchema);

module.exports = Event;
