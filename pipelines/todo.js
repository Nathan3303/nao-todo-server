const Todo = require("../models/todo");
const ObjectId = require("mongoose").Types.ObjectId;
const { makeBoolean } = require("../utils");

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

const handleTagId = (tagId) => {
    return tagId
        ? Todo.aggregate()
              .match({ tags: { $in: [tagId] } })
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

const handleLookupTags = () => {
    return Todo.aggregate()
        .lookup({
            from: "tags",
            localField: "tags",
            foreignField: "_id",
            as: "tagsInfo",
        })
        .pipeline();
};

const handleSelectFields = (fieldsOptions) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: "$_id" },
        projectId: 1,
        project: {
            title: "$_project.title",
        },
        name: 1,
        state: 1,
        priority: 1,
        tags: 1,
        tagsInfo: {
            $map: {
                input: "$tagsInfo",
                as: "tag",
                in: {
                    id: { $toString: "$$tag._id" },
                    name: "$$tag.name",
                    color: "$$tag.color",
                },
            },
        },
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

const handleRelativeDate = (relativeDate) => {
    if (!relativeDate) return [];
    let agg = null;
    switch (relativeDate) {
        case "today":
            agg = Todo.aggregate().match({
                "dueDate.endAt": {
                    $gte: new Date(new Date().setHours(0, 0, 0, 0)),
                    $lte: new Date(new Date().setHours(23, 59, 59, 999)),
                },
            });
            break;
        case "tomorrow":
            agg = Todo.aggregate().match({
                "dueDate.endAt": {
                    $gte: new Date(
                        new Date().setHours(0, 0, 0, 0) + 24 * 60 * 60 * 1000
                    ),
                    $lte: new Date(
                        new Date().setHours(23, 59, 59, 999) +
                            24 * 60 * 60 * 1000
                    ),
                },
            });
            break;
        case "week":
            agg = Todo.aggregate().match({
                "dueDate.endAt": {
                    $gte: new Date(new Date().setHours(0, 0, 0, 0)),
                    $lte: new Date(
                        new Date().setHours(23, 59, 59, 999) +
                            7 * 24 * 60 * 60 * 1000
                    ),
                },
            });
            break;
        default:
            agg = null;
    }
    return agg ? agg.sort({ "dueDate.endAt": 1 }).pipeline() : [];
};

module.exports = {
    handleUserId,
    handleProjectId,
    handleTagId,
    handleId,
    handleName,
    handleState,
    handlePriority,
    handleIsFavorited,
    handleIsDeleted,
    handlePage,
    handleLookupProject,
    handleLookupTags,
    handleSelectFields,
    handleGroupByState,
    handleGroupByPriority,
    handleCountTotal,
    handleRelativeDate,
};
