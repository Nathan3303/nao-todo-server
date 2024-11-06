const Todo = require("../../models/todo");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");

const updateTodos = async (request, response, _p) => {
    const { todoIds, updateInfo } = request.body;

    try {
        const updateRes = await Todo.updateMany(
            { _id: { $in: todoIds } },
            {
                $set: {
                    ...updateInfo,
                    updatedAt: new Date(),
                },
            }
        ).exec();
        if (!updateRes) throw new Error("Update failed");
        const { modifiedCount, matchedCount } = updateRes;
        if (modifiedCount !== matchedCount) throw new Error("Update failed");
        response.status(200).json(buildRD.success("Update successful"));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};

module.exports = updateTodos;
