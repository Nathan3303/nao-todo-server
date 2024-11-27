import {
    getModelForClass,
    index,
    modelOptions,
    mongoose,
    prop
} from '@typegoose/typegoose';

@index({ userId: 1, todoId: 1 })
@index({ userId: 1, todoId: 1, commentId: 1 })
@modelOptions({
    schemaOptions: { timestamps: true, collection: 'comments' },
    options: { allowMixed: 0 }
})
class Comment {
    @prop({ required: true, ref: 'User' })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true, ref: 'Todo' })
    todoId: mongoose.Types.ObjectId;

    @prop({ default: null, ref: 'Comment' })
    commentId: mongoose.Types.ObjectId;

    @prop({ required: true })
    content: string;

    @prop({ default: [] })
    attachments: string[];

    @prop({ default: false })
    isTopUp: boolean;
}

export default getModelForClass(Comment);
export const CommentClass = Comment;
