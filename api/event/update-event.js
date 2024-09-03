const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createTodo(request, response) {
    if (checkMethod(request, response, "PUT")) return;

    const { userId, eventId } = request.query;
    const { title, isDone, isTopped } = request.body;

    if (!eventId) {
        response.status(200).json(buildRD.error("Event id is required"));
        return;
    }

    try {
        const updateResult = await Event.updateOne(
            { userId, _id: new ObjectId(eventId) },
            { $set: { title, isDone, isTopped, updatedAt: Date.now() } }
        );
        if (updateResult && updateResult.modifiedCount) {
            // console.log(updateResult);
            const event = await Event.findById(eventId).select({
                _id: 0,
                id: { $toString: "$_id" },
                title: 1,
                isDone: 1,
                isTopped: 1,
            });
            response.status(200).json(buildRD.success(event));
        } else {
            response.status(200).json(buildRD.error("Event update failed"));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
