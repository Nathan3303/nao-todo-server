const Todo = require("../../models/todo");
const serialExecution = require("../../utils/serial-execution");
const buildRD = require("../../utils/build-response-data");
const { checkMethod, checkQueryLength } = require("../../utils");
const ObjectId = require("mongoose").Types.ObjectId;
const { verifyJWT } = require("../../utils/make-jwt");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "GET")) return;
    if (checkQueryLength(request, response)) return;

    const { authorization } = request.headers;
    if (!authorization) {
        response
            .status(200)
            .json(buildRD.error("Authorization header is missing."));
        return;
    }
    const token = authorization.replace("Bearer ", "");
    const verifyResult = verifyJWT(token);
    if (!verifyResult) {
        response.status(200).json(buildRD.error("Invalid token."));
        return;
    }

    try {
        let basic = await serialExecution([
            () => {
                const { id } = request.query;
                if (!id) return [];
                return Todo.aggregate()
                    .match({ _id: new ObjectId(id) })
                    .pipeline();
            },
            () => {
                const { userId } = request.query;
                if (!userId) return [];
                return Todo.aggregate()
                    .match({ userId: new ObjectId(userId) })
                    .pipeline();
            },
            () => {
                const { projectId } = request.query;
                if (!projectId) return [];
                return Todo.aggregate()
                    .match({ projectId: new ObjectId(projectId) })
                    .pipeline();
            },
            () => {
                const { name } = request.query;
                if (!name) return [];
                return Todo.aggregate()
                    .match({ name: { $regex: name, $options: "i" } })
                    .pipeline();
            },
            () => {
                const { state } = request.query;
                if (!state) return [];
                const splitedState = state.split(",");
                return Todo.aggregate()
                    .match({ state: { $in: splitedState } })
                    .pipeline();
            },
            () => {
                const { priority } = request.query;
                if (!priority) return [];
                const splitedPriority = priority.split(",");
                return Todo.aggregate()
                    .match({ priority: { $in: splitedPriority } })
                    .pipeline();
            },
            () => {
                const { isPinned } = request.query;
                if (!isPinned) return [];
                return Todo.aggregate()
                    .match({ isPinned: !!isPinned })
                    .pipeline();
            },
        ]);
        let query = await serialExecution([
            () => {
                let { page, limit } = request.query;
                page = page || 1;
                limit = parseInt(limit) || 10;
                return Todo.aggregate()
                    .sort("field -createdAt")
                    .skip((page - 1) * limit)
                    .limit(limit)
                    .pipeline();
            },
            () => {
                return Todo.aggregate()
                    .lookup({
                        from: "projects",
                        localField: "projectId",
                        foreignField: "_id",
                        as: "project",
                    })
                    .pipeline();
            },
            () => {
                return Todo.aggregate()
                    .project({
                        _id: 0,
                        id: { $toString: "$_id" },
                        projectId: 1,
                        project: 1,
                        name: 1,
                        state: 1,
                        priority: 1,
                        tags: 1,
                        isDone: 1,
                        createdAt: 1,
                        updatedAt: 1,
                        description: 1,
                        isPinned: 1,
                        dueDate: 1,
                    })
                    .pipeline();
            },
        ]);
        let groupByState = await serialExecution([
            () => {
                const { projectId } = request.query;
                if (!projectId) return [];
                return Todo.aggregate()
                    .match({ projectId: new ObjectId(projectId) })
                    .pipeline();
            },
            () =>
                Todo.aggregate()
                    .group({ _id: "$state", totalTodos: { $sum: 1 } })
                    .pipeline(),
        ]);
        let groupByPriority = await serialExecution([
            () => {
                const { projectId } = request.query;
                if (!projectId) return [];
                return Todo.aggregate()
                    .match({ projectId: new ObjectId(projectId) })
                    .pipeline();
            },
            () =>
                Todo.aggregate()
                    .group({ _id: "$priority", totalTodos: { $sum: 1 } })
                    .pipeline(),
        ]);

        basic = basic.flat();
        query = query.flat();
        groupByState = groupByState.flat();
        groupByPriority = groupByPriority.flat();

        const total = await Todo.aggregate(
            basic.concat([{ $count: "totalCount" }])
        );
        const todos = await Todo.aggregate(basic.concat(query));
        const stateCount = await Todo.aggregate(groupByState);
        const priorityCount = await Todo.aggregate(groupByPriority);

        // console.log(stateCount, priorityCount);

        const _total = total.length ? total[0].totalCount : 0;
        let totalCount = 0;
        const countByState = stateCount.reduce((acc, cur) => {
            const { _id, totalTodos } = cur;
            acc[_id] = totalTodos;
            totalCount += totalTodos;
            return acc;
        }, {});
        const countByPriority = priorityCount.reduce((acc, cur) => {
            const { _id, totalTodos } = cur;
            acc[_id] = totalTodos;
            return acc;
        }, {});

        const payload = {
            countInfo: {
                length: todos.length,
                count: _total,
                total: totalCount,
                byState: countByState,
                byPriority: countByPriority,
            },
            pageInfo: {
                page: parseInt(request.query.page) || 1,
                limit: parseInt(request.query.limit) || 10,
                totalPages: Math.ceil(
                    _total / parseInt(request.query.limit) || 1
                ),
            },
        };

        response.status(200).json(buildRD.success({ todos, payload }));
    } catch (error) {
        response.status(500).json(buildRD.error(error.message));
    }
};
