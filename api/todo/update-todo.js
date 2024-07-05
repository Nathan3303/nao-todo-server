const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { userToken, id } = request.query;
    const { name, description, state, priority, dueDate } = request.body;

    if (!id) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    if (!name) {
        response.status(200).json(buildRD.error("Todo name is required"));
        return;
    }

    try {
        const updateResult = await Todo.updateOne(
            { _id: new ObjectId(id) },
            {
                $set: {
                    name,
                    description,
                    state,
                    priority,
                    dueDate,
                    updatedAt: Date.now(),
                },
            }
        );
        const todo = await Todo.findById(id).select({
            _id: 0,
            id: { $toString: "$_id" },
            description: 1,
            projectId: 1,
            project: 1,
            name: 1,
            state: 1,
            priority: 1,
            tags: 1,
            isDone: 1,
            createdAt: 1,
            updatedAt: 1,
            dueDate: 1,
        });
        if (updateResult && updateResult.modifiedCount) {
            // console.log(updateResult);
            response.status(200).json(buildRD.success(todo));
        } else {
            response.status(200).json(buildRD.error("Todo update failed"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
