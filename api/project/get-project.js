const Project = require("../../models/project");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");

const { matchProjectById } = require("./pipelines");

module.exports = async function getProject(request, response) {
    const { id } = request.query;

    if (!id) {
        response.status(200).json(buildRD.error("Project ID is required"));
        return;
    }

    try {
        const tasks = [
            () => matchProjectById(id),
            () => Project.aggregate().pipeline(),
        ];
        const executeResults = await serialExecution(tasks);
        const projects = await Project.aggregate(executeResults.flat());
        // console.log(projects);
        response.status(200).json(buildRD.success(projects));
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
