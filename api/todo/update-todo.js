const Todo = require("../../models/todo");
const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { userToken, id } = request.query;
    const {
        name,
        description,
        state,
        priority,
        dueDate,
        isPinned,
        isDeleted,
        projectId,
    } = request.body;

    if (!id) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    // if (!name) {
    //     response.status(200).json(buildRD.error("Todo name is required"));
    //     return;
    // }

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
                    isPinned,
                    isDeleted,
                    projectId: new ObjectId(projectId),
                    updatedAt: Date.now(),
                },
            }
        );
        if (updateResult && updateResult.modifiedCount) {
            const todo = await Todo.aggregate()
                .match({ _id: new ObjectId(id) })
                .lookup({
                    from: "projects",
                    localField: "projectId",
                    foreignField: "_id",
                    as: "project",
                })
                .project({
                    _id: 0,
                    id: { $toString: "$_id" },
                    description: 1,
                    projectId: 1,
                    project: {
                        $arrayElemAt: ["$project", 0],
                    },
                    name: 1,
                    state: 1,
                    priority: 1,
                    tags: 1,
                    isDone: 1,
                    createdAt: 1,
                    updatedAt: 1,
                    dueDate: 1,
                    isPinned: 1,
                    isDeleted: 1,
                });
            const events = await Event.find({ todoId: id }).select({
                _id: 0,
                id: { $toString: "$_id" },
                title: 1,
                isDone: 1,
                isTopped: 1,
            });
            // console.log(updateResult);
            if (todo) {
                // console.log(todo);
                todo.events = events || [];
                const data = { ...todo[0], events };
                response.status(200).json(buildRD.success(data));
            }
        } else {
            response.status(200).json(buildRD.error("Todo update failed"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
