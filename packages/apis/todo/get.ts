import { Todo, Event } from '@nao-todo-server/models';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { serialExecute } from '@nao-todo-server/utils';
import { todoPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';
import type { Oid } from '@nao-todo-server/utils';

const getTodo = async (req: Request, res: Response) => {
    try {
        if (!req.query.todoId) throw new Error('参数错误，请求无效');

        const todoId = req.query.todoId as unknown as Oid;

        const getTodoTasks = [
            () => todoPipelines.handleId(todoId),
            () => todoPipelines.handleLookupProject(),
            () => todoPipelines.handleLookupTags(),
            () => todoPipelines.handleSelectFields(),
            () => Todo.aggregate().pipeline()
        ];

        const getTodoTasksExecution = await serialExecute(getTodoTasks);
        const getTodoPipelines = getTodoTasksExecution.flat();
        const todo = (await Todo.aggregate(getTodoPipelines)) as Record<
            string,
            any
        >;

        const events = await Event.find({ todoId }).select({
            _id: 0,
            id: { $toString: '$_id' },
            title: 1,
            isDone: 1,
            isTopped: 1
        });
        if (todo) {
            todo.events = events || [];
            const data = { ...todo[0], events };
            res.json(useSuccessfulResponseData(data));
        }
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
            sort
        } = req.query as unknown as InstanceType<typeof Todo> & {
            tagId: Oid;
            relativeDate: string;
            page: string;
            limit: string;
            sort: string;
        };

        const filterTasks = [
            () => todoPipelines.handleUserId(userId),
            () => todoPipelines.handleProjectId(projectId),
            () => todoPipelines.handleTagId(tagId),
            () => todoPipelines.handleId(id),
            () => todoPipelines.handleName(name),
            () => todoPipelines.handleState(state),
            () => todoPipelines.handlePriority(priority),
            () => todoPipelines.handleIsFavorited(isPinned),
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
        const _limit = parseInt(limit) || 10;
        const pageInfo = {
            page: parseInt(page) || 1,
            limit: _limit,
            totalPages: Math.ceil(count / _limit) || 1
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
