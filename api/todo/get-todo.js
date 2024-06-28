const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "GET")) return;

    const { userToken, id } = request.query;

    if (!id) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
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
        });
        if (todo) {
            // console.log(todo);
            response.status(200).json(buildRD.success(todo));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
