import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteTodo = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.todoId) throw new Error('缺少参数，请求无效');

        const todoId = req.query.todoId as string;

        const deletedTodo = await Todo.findOneAndDelete({
            _id: new ObjectId(todoId)
        }).exec();

        if (!deletedTodo) {
            throw new Error('删除失败，服务器错误');
        }

        return res.json(
            useResponseData(20000, '删除成功', { todoId: deletedTodo._id.toString() })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/deleteTodo] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteTodos = async (req: Request, res: Response) => {};

export { deleteTodo, deleteTodos };
