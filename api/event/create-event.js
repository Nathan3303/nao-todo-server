const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");

const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async function createEvent(request, response) {
    if (checkMethod(request, response, "POST")) return;

    const { todoId, title } = request.body;

    if (!todoId) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }
    if (!title) {
        response.status(200).json(buildRD.error("Event title is required"));
        return;
    }

    try {
        const createResult = await Event.create({
            title,
            todoId: new ObjectId(todoId),
        })
        if (createResult) {
            // console.log(createResult);
            const event = {
                id: createResult._id,
                title: createResult.title,
                isDone: createResult.isDone,
                createdAt: createResult.createdAt,
                updatedAt: createResult.updatedAt,
            }
            response.status(200).json(buildRD.success(event));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
