import { Comment, Event, Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteTodo = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.query.todoId) {
            res.json(useSuccessfulResponseData('缺少参数，请求无效'));
            return;
        }

        // 删除待办任务
        const deletedTodo = await Todo.findOneAndDelete({
            _id: new ObjectId(req.query.todoId as string),
            userId: new ObjectId(userId)
        }).exec();

        // 判断删除结果
        if (!deletedTodo) {
            res.json(useSuccessfulResponseData('删除失败'));
            return;
        }

        // 删除任务下的检查事项和评论
        await Event.deleteMany({
            todoId: new ObjectId(req.query.todoId as string),
            userId: new ObjectId(userId)
        });
        await Comment.deleteMany({
            todoId: new ObjectId(req.query.todoId as string),
            userId: new ObjectId(userId)
        });

        // 返回响应
        return res.json(
            useResponseData(20000, '删除成功', {
                todoId: deletedTodo._id.toString()
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/deleteTodo] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteTodos = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.body.todoIds) {
            res.json(useSuccessfulResponseData('缺少参数，请求无效'));
            return;
        }

        // 删除待办任务
        const deleteResult = await Todo.deleteMany({
            _id: {
                $in: req.body.todoIds.map((id: string) => new ObjectId(id))
            },
            userId: new ObjectId(userId)
        }).exec();

        // 判断删除结果
        if (!deleteResult.deletedCount) {
            res.json(useSuccessfulResponseData('删除失败'));
            return;
        }

        // 删除任务下的检查事项和评论
        await Event.deleteMany({
            todoId: {
                $in: req.body.todoIds.map((id: string) => new ObjectId(id))
            },
            userId: new ObjectId(userId)
        });
        await Comment.deleteMany({
            todoId: {
                $in: req.body.todoIds.map((id: string) => new ObjectId(id))
            },
            userId: new ObjectId(userId)
        });

        // 返回响应
        return res.json(
            useResponseData(20000, '删除成功', {
                todoIds: req.body.todoIds,
                count: deleteResult.deletedCount
            })
        );
    } catch (e) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/todo/deleteTodos] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { deleteTodo, deleteTodos };
