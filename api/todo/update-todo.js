const Todo = require("../../models/todo");
const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;
const getTodo = require("./get-todo");

module.exports = async function updateTodo(request, response, _p) {
    const { todoId } = request.query;

    if (!todoId) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
        const projectId = request.body.projectId;
        delete request.body.projectId;
        const updateInfo = { ...request.body, updatedAt: Date.now() };
        if (projectId) {
            Object.assign(updateInfo, { projectId: new ObjectId(projectId) });
        }
        const updateResult = await Todo.updateOne(
            { _id: new ObjectId(todoId) },
            { $set: updateInfo }
        );
        if (updateResult && updateResult.modifiedCount) {
            await getTodo(request, response, _p);
        } else {
            response.status(200).json(buildRD.error("Todo update failed"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
