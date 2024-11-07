import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId, type Oid, serialExecute } from '@nao-todo-server/utils';
import { eventPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getEvent = async (req: Request, res: Response) => {};

const getEvents = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId) {
            throw new Error('缺少参数，请求无效');
        }

        const { userId, todoId, id, title, isDone, isTopped, page, limit } =
            req.query as unknown as {
                userId: Oid;
                todoId?: Oid;
                id?: Oid;
                title?: string;
                isDone?: boolean;
                isTopped?: boolean;
                page: number;
                limit: number;
            };

        const getEventTasks = [
            () => eventPipelines.handleUserId(userId),
            () => eventPipelines.handleTodoId(todoId),
            () => eventPipelines.handleId(id),
            () => eventPipelines.handleTitle(title),
            () => eventPipelines.handleIsDone(isDone),
            () => eventPipelines.handleIsTopped(isTopped),
            () => eventPipelines.handlePage(page, limit),
            () => eventPipelines.handleSelectFields(),
            () => Event.aggregate().pipeline()
        ];

        const getEventTasksExecution = await serialExecute(getEventTasks);
        const getEventPipelines = getEventTasksExecution.flat();
        const events = await Event.aggregate(getEventPipelines);

        res.json(useSuccessfulResponseData(events));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/event/getEvents] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { getEvent, getEvents };
