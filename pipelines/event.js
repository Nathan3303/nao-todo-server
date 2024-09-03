const Event = require("../models/todo");
const ObjectId = require("mongoose").Types.ObjectId;
const { makeBoolean } = require("../utils");

const handleUserId = (userId) => {
    return userId
        ? Event.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleTodoId = (todoId) => {
    return todoId
        ? Event.aggregate()
              .match({ todoId: new ObjectId(todoId) })
              .pipeline()
        : [];
};

const handleId = (id) => {
    return id
        ? Event.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

const handleTitle = (title) => {
    return title
        ? Event.aggregate()
              .match({ title: { $regex: title, $options: "i" } })
              .pipeline()
        : [];
};

const handleIsDone = (isDone) => {
    return isDone
        ? Event.aggregate()
              .match({ isDone: makeBoolean(isDone) })
              .pipeline()
        : [];
};

const handleIsTopped = (isTopped) => {
    return isTopped
        ? Event.aggregate()
              .match({ isTopped: makeBoolean(isTopped) })
              .pipeline()
        : [];
};

const handlePage = (page, limit) => {
    page = parseInt(page) || 1;
    limit = parseInt(limit) || 10;
    return Event.aggregate()
        .skip((page - 1) * limit)
        .limit(limit)
        .pipeline();
};

const handleSelectFields = (fieldsOptions) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: "$_id" },
        title: 1,
        isDone: 1,
        isTopped: 1,
    };
    return Event.aggregate().project(fieldsOptions).pipeline();
};

module.exports = {
    handleUserId,
    handleTodoId,
    handleId,
    handleTitle,
    handleIsDone,
    handleIsTopped,
    handlePage,
    handleSelectFields,
};
