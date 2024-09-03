const Tag = require("../../models/tag");
const buildRD = require("../../utils/build-response-data");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function deleteTag(request, response, _p) {
    const { userId, tagId } = request.query;

    if (!userId) {
        response.status(200).json(buildRD.error("User id is required"));
        return;
    }
    if (!tagId) {
        response.status(200).json(buildRD.error("Tag id is required"));
        return;
    }

    try {
        const deleteResult = await Tag.findOneAndDelete({
            _id: new ObjectId(tagId),
            userId: new ObjectId(userId),
        });
        response.status(200).json(buildRD.success(deleteResult));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
