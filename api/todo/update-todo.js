const Todo = require("../../models/todo");
const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;
const getTodo = require("./get-todo");

module.exports = async function updateTodo(request, response, _p) {
    const { todoId } = request.query;
    const { projectId } = request.body;

    if (!todoId) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
        const updateResult = await Todo.updateOne(
            { _id: new ObjectId(todoId) },
            {
                $set: {
                    ...request.body,
                    projectId: new ObjectId(projectId),
                    updatedAt: Date.now(),
                },
            }
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
