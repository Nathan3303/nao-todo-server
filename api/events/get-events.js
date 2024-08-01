const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const serialExecution = require("../../utils/serial-execution");

module.exports = async function createTodo(request, response, _p) {
    const { userId, todoId, id, title, isDone, isTopped, page, limit } =
        request.query;

    if (!userId) {
        response.status(200).json(buildRD.error("userId is required"));
        return;
    }

    try {
        const getEventTasks = [
            () => _p.handleUserId(userId),
            () => _p.handleTodoId(todoId),
            () => _p.handleId(id),
            () => _p.handleTitle(title),
            () => _p.handleIsDone(isDone),
            () => _p.handleIsTopped(isTopped),
            () => _p.handlePage(page, limit),
            () => _p.handleSelectFields(),
            () => Event.aggregate().allowDiskUse(true).pipeline(),
        ];
        const getEventTasksExecution = await serialExecution(getEventTasks);
        const getEventPipelines = getEventTasksExecution.flat();
        const events = await Event.aggregate(getEventPipelines);
        if (events) {
            // console.log(events);
            response.status(200).json(buildRD.success(events));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
