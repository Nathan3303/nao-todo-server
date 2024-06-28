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
        const objectId = new ObjectId(id);
        // console.log(objectId);
        const deleteResult = await Todo.deleteOne({ _id: objectId });
        if (deleteResult && deleteResult.deletedCount) {
            // console.log(deleteResult);
            response
                .status(200)
                .json(buildRD.success("Todo deleted successfully"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
