const Project = require("../../models/project");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function updateProject(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { projectId } = request.query;
    const { name, description } = request.body;

    if (!projectId) {
        response.status(200).json(buildRD.error("Project ID is required."));
        return;
    }
    if (name === "") {
        response.status(200).json(buildRD.error("Project name is required."));
        return;
    }

    try {
        const updateRes = await Project.updateOne(
            {
                _id: new ObjectId(projectId),
            },
            {
                $set: {
                    name,
                    description,
                },
            }
        );

        console.log(updateRes);

        response
            .status(200)
            .json(buildRD.success("Project updated successfully."));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
