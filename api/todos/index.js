const Todo = require("../../models/todo");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");
const { checkMethod, checkQueryLength } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async (request, response) => {
    if (checkMethod(request, response, "GET")) return;
    if (checkQueryLength(request, response)) return;

    const tasks = [
        () => {
            const { projectId } = request.query;
            if (!projectId) return [];
            return Todo.aggregate()
                .match({ projectId: new ObjectId(projectId) })
                .pipeline();
        },
        () => {
            const { name } = request.query;
            if (!name) return [];
            return Todo.aggregate()
                .match({ name: { $regex: name, $options: "i" } })
                .pipeline();
        },
        () => {
            const { state } = request.query;
            if (!state) return [];
            return Todo.aggregate().match({ state }).pipeline();
        },
        () => {
            const { parseProjectId } = request.query;
            if (!parseProjectId) return [];
            return Todo.aggregate()
                .lookup({
                    from: "projects",
                    localField: "projectId",
                    foreignField: "_id",
                    as: "project",
                })
                .pipeline();
        },
    ];

    try {
        const executeResults = await serialExecution(tasks);
        const todos = await Todo.aggregate(executeResults.flat());
        response.status(200).json(buildRD.success(todos));
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
