const Tag = require("../../models/tag");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function updateTag(request, response, _p) {
    const { userId, tagId } = request.query;
    const { name, color } = request.body;

    if (!userId) {
        response.status(200).json(buildRD.error("User id is required"));
        return;
    }
    if (!tagId) {
        response.status(200).json(buildRD.error("Tag id is required"));
        return;
    }
    if (!name) {
        response.status(200).json(buildRD.error("Tag name is required"));
        return;
    }

    try {
        const updateResult = await Tag.updateOne(
            {
                _id: new ObjectId(tagId),
                userId: new ObjectId(userId),
            },
            {
                name,
                color,
                updatedAt: new Date(),
            }
        );
        // console.log(updateResult);
        // if (updateResult && updateResult.modifiedCount) {

        // }
        response.status(200).json(buildRD.success(updateResult));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
