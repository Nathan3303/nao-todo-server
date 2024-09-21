const Project = require("../../models/project");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function updateProject(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { projectId } = request.query;
    const { title } = request.body;

    if (!projectId) {
        response.status(200).json(buildRD.error("Project ID is required."));
        return;
    }
    if (title === "") {
        response.status(200).json(buildRD.error("Project title is required."));
        return;
    }

    try {
        const updateRes = await Project.updateOne(
            { _id: new ObjectId(projectId) },
            {
                $set: {
                    ...request.body,
                    updatedAt: Date.now(),
                },
            }
        );
        // console.log(updateRes);
        if (updateRes && updateRes.modifiedCount > 0) {
            response
                .status(200)
                .json(buildRD.success("Project updated successfully."));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
