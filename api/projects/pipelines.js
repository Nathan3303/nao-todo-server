const Project = require("../../models/project");
const ObjectId = require("mongoose").Types.ObjectId;

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
    isDeleted = !!isDeleted;
    return Project.aggregate().match({ isDeleted }).pipeline();
};

const handleIsFinished = (isFinished) => {
    if (isFinished === null || isFinished === undefined) {
        return [];
    }
    const _bool = new Boolean(isFinished);
    return Project.aggregate().match({ isFinished: _bool }).pipeline();
};

const handlePage = (page, limit) => {
    page = page || 1;
    limit = limit || 10;
    const skip = (page - 1) * limit;
    return Project.aggregate().skip(skip).limit(limit).pipeline();
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
        })
        .pipeline();
};

module.exports = {
    handleUserId,
    handleTitle,
    handleIsDeleted,
    handleIsFinished,
    handlePage,
    handleSort,
    handleOutput,
};
