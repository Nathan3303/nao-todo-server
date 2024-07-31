const Tag = require("../../models/tag");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    const { userId, name, color } = request.body;

    if (!userId) {
        response.status(200).json(buildRD.error("userId is required"));
        return;
    }
    if (!name) {
        response.status(200).json(buildRD.error("name is required"));
        return;
    }
    if (!color) {
        response.status(200).json(buildRD.error("color is required"));
        return;
    }

    try {
        const createResult = await Tag.create({
            userId: new ObjectId(userId),
            name,
            color,
        });
        if (createResult) {
            // console.log(createResult);
            const responseData = {
                id: createResult._id,
                name: createResult.name,
                color: createResult.color,
                createdAt: createResult.createdAt,
                updatedAt: createResult.updatedAt,
            };
            response.status(200).json(buildRD.success(responseData));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
