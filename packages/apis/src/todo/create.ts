import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';
import moment from 'moment';

const createTodo = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId) throw new Error('用户ID不能为空');
        if (!req.body.projectId) throw new Error('项目ID不能为空');
        if (!req.body.name) throw new Error('标题不能为空');

        const now = moment();

        const todo = await Todo.create({
            userId,
            projectId: req.body.projectId,
            name: req.body.name,
            description: req.body.description,
            state: req.body.state,
            priority: req.body.priority,
            dueDate: {
                startAt: now.startOf('day').toISOString(true),
                ...req.body.dueDate
            },
            tags: req.body.tags
        });

        if (!todo) throw new Error('创建失败');

        return res.json(
            useSuccessfulResponseData({
                ...todo.toJSON(),
                id: todo._id.toString()
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/createTodo] Error:', e);
        return;
    }
};

const createTodos = async (req: Request, res: Response) => {};

export { createTodo, createTodos };
