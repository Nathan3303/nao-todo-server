const mongoose = require("mongoose");
const Project = require("../../models/project");
const serialExecution = require("../../utils/serial-execution");
const buildResponseData = require("../../utils/build-response-data");
const { checkMethod, checkQueryLength } = require("../../utils");
const ObjectId = mongoose.Types.ObjectId;

module.exports = async (request, response) => {
    if (checkMethod(request, response, "GET")) return;
    if (checkQueryLength(request, response)) return;

    const tasks = [
        () => {
            const { userId } = request.query;
            if (!userId) return [];
            return Project.aggregate()
                .match({ owner: new ObjectId(userId) })
                .sort({ created_at: -1 })
                .pipeline();
        },
        () => {
            return Project.aggregate()
                .project({
                    _id: 0,
                    id: { $toString: "$_id" },
                    name: 1,
                    description: 1,
                    created_at: 1,
                    updated_at: 1,
                })
                .pipeline();
        },
    ];

    try {
        const executeResults = await serialExecution(tasks);
        const projects = await Project.aggregate(executeResults.flat());
        response.status(200).json(buildResponseData("20000", "OK", projects));
    } catch (error) {
        response
            .status(500)
            .json(buildResponseData("50000", "Not OK", error.message));
    }
};
