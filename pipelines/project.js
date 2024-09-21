const Project = require("../models/project");
const ObjectId = require("mongoose").Types.ObjectId;
const { makeBoolean } = require("../utils");

const handleUserId = (userId) => {
    return userId
        ? Project.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleTitle = (title) => {
    return title
        ? Project.aggregate()
              .match({ title: { $regex: title, $options: "i" } })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted) => {
    if (isDeleted === null || isDeleted === undefined) {
        return [];
    }
    const _bool = makeBoolean(isDeleted);
    return Project.aggregate().match({ isDeleted: _bool }).pipeline();
};

const handleIsFinished = (isFinished) => {
    if (isFinished === null || isFinished === undefined) {
        return [];
    }
    const _bool = makeBoolean(isFinished);
    return Project.aggregate().match({ isFinished: _bool }).pipeline();
};

const handleIsArchived = (isArchived) => {
    if (isArchived === null || isArchived === undefined) {
        return [];
    }
    const _bool = makeBoolean(isArchived);
    return Project.aggregate().match({ isArchived: _bool }).pipeline();
};

const handlePage = (page, limit) => {
    page = parseInt(page) || 1;
    limit = parseInt(limit) || 10;
    return Project.aggregate()
        .skip((page - 1) * limit)
        .limit(limit)
        .pipeline();
};

const handleSort = (sort) => {
    if (sort === null || sort === undefined) {
        return [];
    }
    const sortObject = {};
    sort = sort.split(",");
    sort.forEach((s) => {
        if (s.startsWith("-")) {
            s = s.slice(1);
            sortObject[s] = -1;
        } else {
            sortObject[s] = 1;
        }
    });
    return Project.aggregate().sort(sortObject).pipeline();
};

const handleOutput = () => {
    return Project.aggregate()
        .project({
            _id: 0,
            id: { $toString: "$_id" },
            userId: 1,
            title: 1,
            description: 1,
            state: 1,
            isArchived: 1,
            isFavorite: 1,
            isDeleted: 1,
            isFinished: 1,
            createdAt: 1,
            updatedAt: 1,
            deletedAt: 1,
            finishedAt: 1,
            archivedAt: 1,
            viewType: 1,
            defaultFilterInfo: 1,
            defaultSortInfo: 1,
            defaultColumns: 1,
        })
        .pipeline();
};

module.exports = {
    handleUserId,
    handleTitle,
    handleIsDeleted,
    handleIsFinished,
    handleIsArchived,
    handlePage,
    handleSort,
    handleOutput,
};
