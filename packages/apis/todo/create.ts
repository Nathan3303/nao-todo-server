import { Todo } from '@nao-todo-server/models';
import { useErrorResponseData, useResponseData } from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const createTodo = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId) throw new Error('用户ID不能为空');
        if (!req.body.projectId) throw new Error('项目ID不能为空');
        if (!req.body.name) throw new Error('标题不能为空');

        const todo = await Todo.create({
            userId: req.body.userId,
            projectId: req.body.projectId,
            name: req.body.name,
            description: req.body.description,
            state: req.body.state,
            priority: req.body.priority,
            dueDate: req.body.dueDate,
            tags: req.body.tags
        });

        if (!todo) throw new Error('创建失败');

        return res.json(
            useResponseData(20000, '创建成功', { todoId: todo._id })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/create] Error:', e);
        return;
    }
};

const createTodos = async (req: Request, res: Response) => {};

export { createTodo, createTodos };
