import { Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const updateTag = async (req: Request, res: Response) => {
    try {
        if (!req.query.userId || !req.query.tagId) {
            throw new Error('参数错误，请求无效');
        }

        const { userId, tagId } = req.query as {
            userId: string;
            tagId: string;
        };

        const updatedTag = await Tag.findOneAndUpdate(
            {
                _id: new ObjectId(tagId),
                userId: new ObjectId(userId)
            },
            {
                $set: {
                    ...req.body,
                    updatedAt: new Date()
                }
            }
        );

        if (!updatedTag) throw new Error('标签更新失败');

        return res.json(useSuccessfulResponseData(updatedTag?._id || null));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/updateTag] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const updateTags = async (req: Request, res: Response) => {};

export { updateTag, updateTags };
