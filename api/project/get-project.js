const Project = require("../../models/project");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");
const { checkMethod, checkQueryLength } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function getProject(request, response) {
    if (checkMethod(request, response, "GET")) return;
    if (checkQueryLength(request, response)) return;

    const tasks = [
        () => {
            const { id, name } = request.query;
            if (id) {
                return Project.aggregate()
                    .match({ _id: new ObjectId(id) })
                    .pipeline();
            }
            if (name) {
                return Project.aggregate()
                    .match({ name: { $regex: name, $options: "i" } })
                    .pipeline();
            }
            return [];
        },
    ];

    try {
        const executeResults = await serialExecution(tasks);
        const projects = await Project.aggregate(executeResults.flat());
        response.status(200).json(buildRD.success(projects));
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
