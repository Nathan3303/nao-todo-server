const mongoose = require("mongoose");

const ObjectId = mongoose.Schema.Types.ObjectId;

const tagSchema = new mongoose.Schema({
    userId: { type: ObjectId, ref: "User" },
    name: { type: String, required: true },
    color: { type: String, required: true },
    isDeleted: { type: Boolean, default: false },
    deletedAt: { type: Date },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now },
});

const Tag = mongoose.model("Tag", tagSchema);

module.exports = Tag;
