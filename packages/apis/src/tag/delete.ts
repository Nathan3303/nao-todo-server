import { Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const deleteTag = async (req: Request, res: Response) => {
    try {
        if (!req.query.userId || !req.query.tagId) {
            throw new Error('参数错误，请求无效');
        }

        const { userId, tagId } = req.query as {
            userId: string;
            tagId: string;
        };

        const deletedTag = await Tag.findOneAndDelete({
            _id: new ObjectId(tagId),
            userId: new ObjectId(userId)
        });

        return res.json(useSuccessfulResponseData(deletedTag?._id || null));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/deleteTag] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteTags = async (req: Request, res: Response) => {};

export { deleteTag, deleteTags };
