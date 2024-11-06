const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../../utils");

const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createEvent(request, response) {
    if (checkMethod(request, response, "DELETE")) return;

    const { userId, eventId } = request.query;
    // console.log(id)

    if (!eventId) {
        response.status(200).json(buildRD.error("Event id is required"));
        return;
    }

    try {
        const deleteResult = await Event.findOneAndDelete({
            userId,
            _id: new ObjectId(eventId),
        });
        if (deleteResult) {
            console.log(deleteResult);
            response.status(200).json(buildRD.success(true));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
