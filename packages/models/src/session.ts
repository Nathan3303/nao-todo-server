import {
    getModelForClass,
    index,
    modelOptions,
    mongoose,
    prop
} from '@typegoose/typegoose';

@index({ userId: 1 })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'sessions' } })
class Session {
    @prop({ required: true, type: mongoose.Schema.Types.ObjectId })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true })
    token: string;

    @prop({ required: true })
    expiresAt: Date;
}

export default getModelForClass(Session);
export { Session };
