const Project = require("../../models/project");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createProject(request, response) {
    const { userId, title, description } = request.body;

    if (!userId) {
        response.status(200).json(buildRD.error("User id is required"));
        return;
    }
    if (!title) {
        response.status(200).json(buildRD.error("Project title is required"));
        return;
    }

    try {
        // Check if project with same name already exists
        const isExists = await Project.findOne({
            title,
            userId,
        }).countDocuments();
        // console.log(project);
        if (isExists) {
            response
                .status(200)
                .json(buildRD.error("Project with same name already exists"));
            return;
        }
        // Create new project
        const createRes = await Project.create({
            title,
            description,
            userId: new ObjectId(userId),
        });
        if (!createRes) {
            response
                .status(200)
                .json(buildRD.error("Failed to create project"));
            return;
        }
        response
            .status(200)
            .json(buildRD.success({ id: createRes._id, ...createRes._doc }));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
