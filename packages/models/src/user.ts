import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@index({ email: 1, password: 1 })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'users' } })
class User {
    @prop({ required: true, unique: true })
    username: string;

    @prop({ required: true })
    password: string;

    @prop({ required: true, unique: true })
    email: string;

    @prop()
    nickName: string;

    @prop()
    avatar: string;

    @prop({ required: true, default: 'user' })
    role: string;
}

export default getModelForClass(User);
export { User };
