const Project = require("../../models/project");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createProject(request, response) {
    if (checkMethod(request, response, "POST")) return;

    const { userId, name, description } = request.body;
    if (!userId) {
        response.status(200).json(buildRD.error("User id is required"));
        return;
    }
    if (!name) {
        response.status(200).json(buildRD.error("Name is required"));
        return;
    }

    try {
        const project = await Project.findOne({
            name,
            owner: new ObjectId(userId),
        });
        if (project) {
            response.status(200).json(buildRD.error("Project already exists"));
            return;
        }
        const createRes = await Project.create({
            name,
            description,
            owner: new ObjectId(userId),
        });
        if (!createRes) {
            response
                .status(200)
                .json(buildRD.error("Failed to create project"));
            return;
        }
        response.status(200).json(
            buildRD.success({
                id: createRes._id,
                name: createRes.name,
                description: createRes.description,
                owner: createRes.owner,
                createdAt: createRes.created_at,
                updatedAt: createRes.updated_at,
            })
        );
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
        return;
    }
};
