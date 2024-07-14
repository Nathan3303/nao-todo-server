const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "POST")) return;

    const { userId, projectId, name } = request.body;

    if (!userId) {
        response.status(200).json(buildRD.error("userId is required"));
        return;
    }

    if (!projectId) {
        response.status(200).json(buildRD.error("projectId is required"));
        return;
    }
    if (!name) {
        response.status(200).json(buildRD.error("name is required"));
        return;
    }

    try {
        const createResult = await Todo.create({ name, projectId, userId });
        if (createResult) {
            console.log(createResult);
            response.status(200).json(buildRD.success(createResult));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
