const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function deleteTodo(request, response) {
    if (checkMethod(request, response, "DELETE")) return;

    const { todoId } = request.query;

    if (!todoId) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
        const _id = new ObjectId(todoId);
        // const updateResult = await Todo.updateOne(
        //     { _id },
        //     { $set: { isDeleted: true } }
        // );
        // if (updateResult && updateResult.modifiedCount) {
        //     const todo = await Todo.findOne({ _id });
        //     if (todo) {
        //         response.status(200).json(buildRD.success(todo));
        //         return;
        //     }
        //     throw new Error("Todo not found");
        // }
        // throw new Error("Todo not found");
        const deleteResult = await Todo.findOneAndDelete({ _id });
        // console.log(deleteResult)
        if (deleteResult && deleteResult._id) {
            response
                .status(200)
                .json(buildRD.success({ id: deleteResult._id }));
            return;
        }
        throw new Error("Todo not found");
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
