const Todo = require("../../models/todo");
const ObjectId = require("mongoose").Types.ObjectId;

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

module.exports = {
    handleUserId,
    handleProjectId,
};
