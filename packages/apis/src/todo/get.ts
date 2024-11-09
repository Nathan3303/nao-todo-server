import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { serialExecute } from '@nao-todo-server/utils';
import { todoPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getTodo = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.todoId) {
            throw new Error('参数错误，请求无效');
        }

        const todoId = req.query.todoId as string;

        const getTodoTasks = [
            () => todoPipelines.handleId(todoId),
            () => todoPipelines.handleLookupProject(),
            () => todoPipelines.handleLookupTags(),
            () => todoPipelines.handleSelectFields(),
            () => Todo.aggregate().pipeline()
        ];

        const getTodoTasksExecution = await serialExecute(getTodoTasks);
        const getTodoPipelines = getTodoTasksExecution.flat();

        const todo = (await Todo.aggregate(getTodoPipelines).exec()) || {};

        res.json(useSuccessfulResponseData(todo));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/getTodo] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const getTodos = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId) throw new Error('参数错误，请求无效');

        const {
            projectId,
            name,
            state,
            priority,
            isFavorited,
            isDeleted,
            relativeDate,
            page,
            limit,
            sort
        } = req.query as unknown as {
            projectId?: string;
            name?: string;
            state?: string;
            priority?: string;
            isFavorited?: boolean;
            isDeleted?: boolean;
            relativeDate?: string;
            sort?: string;
            page: number;
            limit: number;
        };

        const filterTasks = [
            () => todoPipelines.handleUserId(userId),
            () => todoPipelines.handleProjectId(projectId),
            // () => todoPipelines.handleTagId(tagId),
            // () => todoPipelines.handleId(id),
            () => todoPipelines.handleName(name),
            () => todoPipelines.handleState(state),
            () => todoPipelines.handlePriority(priority),
            () => todoPipelines.handleIsFavorited(isFavorited),
            () => todoPipelines.handleIsDeleted(isDeleted),
            () => todoPipelines.handleRelativeDate(relativeDate)
        ];
        const basicTasks = [
            () => todoPipelines.handleLookupProject(),
            () => todoPipelines.handleSelectFields(),
            () => todoPipelines.handleSort(sort),
            () => todoPipelines.handlePage(page, limit)
        ];
        const stateCountTasks = [() => todoPipelines.handleGroupByState()];
        const priorityCountTasks = [
            () => todoPipelines.handleGroupByPriority()
        ];
        const totalCountTasks = [() => todoPipelines.handleCountTotal()];

        const filterTasksExecution = await serialExecute(filterTasks);
        const basicTasksExecution = await serialExecute(basicTasks);
        const stateCountExecution = await serialExecute(stateCountTasks);
        const priorityCountExecution = await serialExecute(priorityCountTasks);
        const totalCountExecution = await serialExecute(totalCountTasks);

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
            byPriority: byPriority
        };
        const pageInfo = {
            page: page || 1,
            limit: limit,
            totalPages: Math.ceil(count / limit) || 1
        };

        const reponseData = { todos, payload: { countInfo, pageInfo } };
        res.json(useSuccessfulResponseData(reponseData));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/getTodos] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { getTodo, getTodos };
