const Tag = require("../models/tag");
const ObjectId = require("mongoose").Types.ObjectId;
const { makeBoolean } = require("../utils");

const handleUserId = (userId) => {
    return userId
        ? Tag.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleName = (name) => {
    return name
        ? Tag.aggregate()
              .match({ name: { $regex: `.*${name}.*`, $options: "i" } })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted) => {
    return isDeleted
        ? Tag.aggregate()
              .match({ isDeleted: makeBoolean(isDeleted) })
              .pipeline()
        : [];
};

const handlePage = (page, limit) => {
    page = parseInt(page) || 1;
    limit = parseInt(limit) || 10;
    return Tag.aggregate()
        .skip((page - 1) * limit)
        .limit(limit)
        .pipeline();
};

const handleSelectFields = (fieldsOptions) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: "$_id" },
        name: 1,
        color: 1,
        createdAt: 1,
        updatedAt: 1,
    };
    return Tag.aggregate().project(fieldsOptions).pipeline();
};

module.exports = {
    handleUserId,
    handleName,
    handleIsDeleted,
    handlePage,
    handleSelectFields,
};
