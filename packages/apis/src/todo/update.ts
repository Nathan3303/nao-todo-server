import { Todo } from '@nao-todo-server/models';
import {
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateTodo = async (req: Request, res: Response) => {
    try {
        if (!req.query.todoId) throw new Error('参数不匹配，请求无效');

        const todoId = req.query.todoId as string;
        const projectId = req.body.projectId as string;
        const updateOptions = { ...req.body, updatedAt: new Date() };
        updateOptions.projectId = new ObjectId(projectId);

        const updatedTodo = await Todo.findOneAndUpdate(
            { _id: new ObjectId(todoId) },
            { $set: updateOptions }
        );
        if (!updatedTodo) throw new Error('更新失败');

        return res.json(
            useResponseData(20000, '更新成功', { todoId: updatedTodo._id })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/updateTodo] Error', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const updateTodos = async (req: Request, res: Response) => {
    try {
        const { todoIds, updateInfo } = req.body;

        const updateRes = await Todo.updateMany(
            { _id: { $in: todoIds } },
            { $set: { ...updateInfo, updatedAt: new Date() } }
        ).exec();
        if (!updateRes) throw new Error('Update failed');

        const { modifiedCount, matchedCount } = updateRes;
        if (modifiedCount !== matchedCount) throw new Error('Update failed');

        res.json(useSuccessfulResponseData('Update successful'));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/updateTodos] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { updateTodo, updateTodos };
