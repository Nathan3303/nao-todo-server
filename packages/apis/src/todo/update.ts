import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateTodo = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.todoId) {
            throw new Error('参数错误，请求无效');
        }

        const todoId = req.query.todoId as string;

        const updatedTodo = await Todo.findOneAndUpdate(
            { _id: new ObjectId(todoId), userId },
            { $set: { ...req.body } },
            { new: true }
        ).exec();

        if (!updatedTodo) {
            throw new Error('更新失败');
        }

        return res.json(
            useResponseData(20000, '更新成功', {
                todoId: updatedTodo._id.toString()
            })
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
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.body.todoIds || !req.body.updateInfo) {
            throw new Error('参数错误，请求无效');
        }

        const { todoIds, updateInfo } = req.body as unknown as {
            todoIds: string[];
            updateInfo: Record<string, unknown>;
        };

        const updateRes = await Todo.updateMany(
            { _id: { $in: todoIds } },
            { $set: { ...updateInfo } },
            { multi: true }
        ).exec();

        if (!updateRes || updateRes.modifiedCount !== updateRes.matchedCount) {
            throw new Error('更新失败');
        }

        res.json(
            useSuccessfulResponseData({
                todoIds,
                total: todoIds.length,
                modifiedCount: updateRes.modifiedCount
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/updateTodos] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { updateTodo, updateTodos };
