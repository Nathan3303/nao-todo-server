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
            let { page, limit } = request.query;
            page = page || 1;
            limit = parseInt(limit) || 10;
            return Todo.aggregate()
                .sort("field -createdAt")
                .skip((page - 1) * limit)
                .limit(limit)
                .pipeline();
        },
        () => {
            return Todo.aggregate()
                .lookup({
                    from: "projects",
                    localField: "projectId",
                    foreignField: "_id",
                    as: "project",
                })
                .pipeline();
        },
        () => {
            return Todo.aggregate()
                .project({
                    _id: 0,
                    id: { $toString: "$_id" },
                    projectId: 1,
                    project: 1,
                    name: 1,
                    state: 1,
                    priority: 1,
                    tags: 1,
                    isDone: 1,
                    createdAt: 1,
                    updatedAt: 1,
                })
                .pipeline();
        },
    ];

    try {
        const executeResults = await serialExecution(tasks);
        const todos = await Todo.aggregate(executeResults.flat());
        const totalDocuments = await Todo.find({
            projectId: new ObjectId(request.query.projectId),
        })
            .countDocuments()
            .exec();
        const totalPages = Math.ceil(
            totalDocuments / parseInt(request.query.limit) || 1
        );
        response.status(200).json(
            buildRD.success({
                todos,
                payload: {
                    count: todos.length,
                    total: totalDocuments,
                    page: parseInt(request.query.page) || 1,
                    limit: parseInt(request.query.limit) || 10,
                    totalPages,
                },
            })
        );
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
