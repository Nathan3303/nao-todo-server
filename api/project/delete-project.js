const Project = require("../../models/project");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");
const { checkMethod, checkQueryLength } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function deleteProject(request, response) {
    if (checkMethod(request, response, "DELETE")) return;
    if (checkQueryLength(request, response)) return;

    const { projectId: id } = request.query;

    if (!id) {
        response.status(200).json(buildRD.error("Project ID is required."));
        return;
    }

    try {
        const project = await Project.findByIdAndDelete(id);
        if (!project) {
            response.status(200).json(buildRD.error("Project not found."));
            return;
        }
        response
            .status(200)
            .json(buildRD.success("Project deleted successfully."));
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
