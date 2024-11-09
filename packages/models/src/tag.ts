import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { timestamps: true, collection: 'tags' } })
class Tag {
    @prop({ required: true })
    userId: string;

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
