const Todo = require("../../models/todo");
const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;
const getTodo = require("./get-todo");

module.exports = async function createTodo(request, response, _p) {
    const { id } = request.query;
    const {
        name,
        description,
        state,
        priority,
        dueDate,
        isDone,
        isPinned,
        isDeleted,
        projectId,
        tags,
    } = request.body;

    if (!id) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    const _tags = tags.map((tagIdString) => {
        tagIdString = tagIdString.trim();
        return new ObjectId(tagIdString);
    });

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
                    isDone,
                    isPinned,
                    isDeleted,
                    projectId: new ObjectId(projectId),
                    tags: _tags,
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
