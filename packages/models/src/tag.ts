import {
    getModelForClass,
    prop,
    modelOptions,
    mongoose
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { timestamps: true, collection: 'tags' } })
class Tag {
    @prop({ required: true, ref: 'User' })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true })
    name: string;

    @prop({ required: true })
    color: string;

    @prop({ default: '' })
    description: string;

    @prop({ default: false })
    isDeleted: boolean;

    @prop({ default: null })
    deletedAt: Date;
}

export default getModelForClass(Tag);
export { Tag };
