const Todo = require("../../models/todo");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");

const getTodos = async (request, response, _p) => {
    const {
        userId,
        projectId,
        tagId,
        id,
        name,
        state,
        priority,
        isPinned,
        isDeleted,
        relativeDate,
        page,
        limit,
        sort,
    } = request.query;

    try {
        const filterTasks = [
            () => _p.handleUserId(userId),
            () => _p.handleProjectId(projectId),
            () => _p.handleTagId(tagId),
            () => _p.handleId(id),
            () => _p.handleName(name),
            () => _p.handleState(state),
            () => _p.handlePriority(priority),
            () => _p.handleIsFavorited(isPinned),
            () => _p.handleIsDeleted(isDeleted),
            () => _p.handleRelativeDate(relativeDate),
        ];
        const basicTasks = [
            () => _p.handleLookupProject(),
            () => _p.handleSelectFields(),
            () => _p.handleSort(sort),
            () => _p.handlePage(page, limit),
        ];
        const stateCountTasks = [() => _p.handleGroupByState()];
        const priorityCountTasks = [() => _p.handleGroupByPriority()];
        const totalCountTasks = [() => _p.handleCountTotal()];

        const filterTasksExecution = await serialExecution(filterTasks);
        const basicTasksExecution = await serialExecution(basicTasks);
        const stateCountExecution = await serialExecution(stateCountTasks);
        const priorityCountExecution = await serialExecution(
            priorityCountTasks
        );
        const totalCountExecution = await serialExecution(totalCountTasks);

        const filterPipelines = filterTasksExecution.flat();
        const basicPipelines = basicTasksExecution.flat();
        const stateCountPipelines = stateCountExecution.flat();
        const priorityCountPipelines = priorityCountExecution.flat();
        const totalCountPipelines = totalCountExecution.flat();

        const todos = await Todo.aggregate(
            filterPipelines.concat(basicPipelines)
        );
        const stateCount = await Todo.aggregate(
            filterPipelines.concat(stateCountPipelines)
        );
        const priorityCount = await Todo.aggregate(
            filterPipelines.concat(priorityCountPipelines)
        );
        const totalCount = await Todo.aggregate(totalCountPipelines);

        let count = 0;
        const byState = stateCount.reduce((acc, cur) => {
            acc[cur._id] = cur.total;
            count += cur.total;
            return acc;
        }, {});
        const byPriority = priorityCount.reduce((acc, cur) => {
            acc[cur._id] = cur.total;
            return acc;
        }, {});

        const countInfo = {
            length: todos.length,
            count,
            total: totalCount[0]?.total || 0,
            byState: byState,
            byPriority: byPriority,
        };
        const _limit = parseInt(limit) || 10;
        const pageInfo = {
            page: parseInt(page) || 1,
            limit: _limit,
            totalPages: Math.ceil(count / _limit) || 1,
        };

        const reponseData = { todos, payload: { countInfo, pageInfo } };
        // console.log(reponseData);
        response.status(200).json(buildRD.success(reponseData));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};

module.exports = getTodos;
