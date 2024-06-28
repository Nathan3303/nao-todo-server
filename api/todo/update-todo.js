const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { userToken, id } = request.query;
    const { name, description, state, priority } = request.body;

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
                    updatedAt: new Date(),
                },
            }
        );
        if (updateResult && updateResult.modifiedCount) {
            // console.log(updateResult);
            response
                .status(200)
                .json(buildRD.success("Todo updated successfully"));
        } else {
            response.status(200).json(buildRD.error("Todo update failed"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
