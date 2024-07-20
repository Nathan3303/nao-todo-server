const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "DELETE")) return;

    const { userToken, id } = request.query;

    if (!id) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
        const todoId = new ObjectId(id);
        const updateResult = await Todo.updateOne(
            { _id: todoId },
            { $set: { isDeleted: true } }
        );
        if (updateResult && updateResult.modifiedCount) {
            const todo = await Todo.findOne({ _id: todoId });
            if (todo) {
                response.status(200).json(buildRD.success(todo));
                return;
            }
            throw new Error("Todo not found");
        }
        throw new Error("Todo not found");
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
