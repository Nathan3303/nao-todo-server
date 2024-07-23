const Todo = require("../../models/todo");
const ObjectId = require("mongoose").Types.ObjectId;
const { makeBoolean } = require("../../utils");

const handleUserId = (userId) => {
    return userId
        ? Todo.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleProjectId = (projectId) => {
    return projectId
        ? Todo.aggregate()
              .match({ projectId: new ObjectId(projectId) })
              .pipeline()
        : [];
};

const handleId = (id) => {
    return id
        ? Todo.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

const handleName = (name) => {
    return name
        ? Todo.aggregate()
              .match({ name: { $regex: name, $options: "i" } })
              .pipeline()
        : [];
};

const handleState = (state) => {
    if (!state) return [];
    const splitedState = state.split(",");
    return Todo.aggregate()
        .match({ state: { $in: splitedState } })
        .pipeline();
};

const handlePriority = (priority) => {
    if (!priority) return [];
    const splitedPriority = priority.split(",");
    return Todo.aggregate()
        .match({ priority: { $in: splitedPriority } })
        .pipeline();
};

const handleIsFavorited = (isFavorited) => {
    return isFavorited
        ? Todo.aggregate()
              .match({ isPinned: makeBoolean(isFavorited) })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted) => {
    return isDeleted
        ? Todo.aggregate()
              .match({ isDeleted: makeBoolean(isDeleted) })
              .pipeline()
        : [];
};

const handlePage = (page, limit) => {
    page = parseInt(page) || 1;
    limit = parseInt(limit) || 10;
    console.log("page", page, "limit", limit)
    return Todo.aggregate()
        .skip((page - 1) * limit)
        .limit(limit)
        .pipeline();
};

const handleLookupProject = () => {
    return Todo.aggregate()
        .lookup({
            from: "projects",
            localField: "projectId",
            foreignField: "_id",
            as: "_project",
        })
        .unwind({ path: "$_project", preserveNullAndEmptyArrays: true })
        .pipeline();
};

const handleSelectFields = (fieldsOptions) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: "$_id" },
        projectId: 1,
        project: {
            // id: { $toString: "$_project._id" },
            title: "$_project.title",
        },
        name: 1,
        state: 1,
        priority: 1,
        tags: 1,
        isDone: 1,
        createdAt: 1,
        updatedAt: 1,
        description: 1,
        isPinned: 1,
        dueDate: 1,
        isDeleted: 1,
    };
    return Todo.aggregate().project(fieldsOptions).pipeline();
};

const handleGroupByState = () => {
    return Todo.aggregate()
        .group({
            _id: "$state",
            total: { $sum: 1 },
            totalTodos: { $sum: 1 },
        })
        .pipeline();
};

const handleGroupByPriority = () => {
    return Todo.aggregate()
        .group({
            _id: "$priority",
            total: { $sum: 1 },
            totalTodos: { $sum: 1 },
        })
        .pipeline();
};

const handleCountTotal = () => {
    return Todo.aggregate().count("total").pipeline();
};

module.exports = {
    handleUserId,
    handleProjectId,
    handleId,
    handleName,
    handleState,
    handlePriority,
    handleIsFavorited,
    handleIsDeleted,
    handlePage,
    handleLookupProject,
    handleSelectFields,
    handleGroupByState,
    handleGroupByPriority,
    handleCountTotal,
};
