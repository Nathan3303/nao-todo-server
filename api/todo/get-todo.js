const Todo = require("../../models/todo");
const Event = require("../../models/event");
const buildRD = require("../../utils/build-response-data");
const serialExecution = require("../../utils/serial-execution");

module.exports = async function createTodo(request, response, _p) {
    const { todoId } = request.query;

    if (!todoId) {
        response.status(200).json(buildRD.error("Todo id is required"));
        return;
    }

    try {
        const getTodoTasks = [
            () => _p.handleId(todoId),
            () => _p.handleLookupProject(),
            () => _p.handleLookupTags(),
            () => _p.handleSelectFields(),
            () => Todo.aggregate().pipeline(),
        ];
        const getTodoTasksExecution = await serialExecution(getTodoTasks);
        const getTodoPipelines = getTodoTasksExecution.flat();
        const todo = await Todo.aggregate(getTodoPipelines);

        const events = await Event.find({ todoId })
            .select({
                _id: 0,
                id: { $toString: "$_id" },
                title: 1,
                isDone: 1,
                isTopped: 1,
            });
        if (todo) {
            // console.log(todo);
            todo.events = events || [];
            const data = { ...todo[0], events };
            response.status(200).json(buildRD.success(data));
        }
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
