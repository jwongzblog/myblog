最近两年的工作都在捣腾openstack和k8ts，这两兄弟可以说是目前人类最先进的开源资源调度系统了，参考我之前关于[算法与算力](https://github.com/jwongzblog/myblog/blob/master/%E4%BA%A7%E5%93%81%E7%9A%84%E6%80%9D%E8%80%83/%E7%AE%97%E6%B3%95%E5%92%8C%E7%AE%97%E5%8A%9B.md)的思考。业余时间开始学习机器学习算法，进展很慢，因为那些晦涩的公式符号难以理解，更何况去理解它的推导了。之所以写这篇文章，因为最近换了一种学习思路，我跳过了这些公式，阅读每个算法想去解决什么问题？用什么方式解决这个问题？刷完一轮后揭开了机器学习的神秘面纱，不过略微有点失望

## 机器学习的本质是什么
MachineLearning基本上就是解决分类问题，大神们先别打脸，这里我把拟合和线性回归，离群点分析也称作分类问题，我称为在线和不在线上。而所有算法的解决思路都是人为标注，工程师给定一个正确答案，机器学习通过各种算法让原始数据算出一个特征，或者原始数据本身就自带特征，工程师再通过算法尝试搞清楚这些特征在二维或者多维坐标中的分布情况，深度学习的结果也是分类，我在这里分为正确、接近正确、不正确。增强学习更是需要人类反馈正确答案，算法本身再去自我调整。机器学习试图通过计算手段，利用人类经验来提炼算法和模型，通过计算来解决相似问题。

## 如果数据需要标注，意味着什么
人类知道规则和结果的好坏，AI不知道，如果AI自我进化的过程中不知道什么结果是好什么是坏，它又会进化成什么样子？混沌但最终毫无用处？还是混沌到非常可怕，朝着人类无法驾驭的坏方向发展？所以AI真的能反应人类的思维吗？人类的思考或者行为确实有些规律可寻，但人类的思维本质上是随机的吗？或者人类自认为是随机的，但本质是趋利的，有迹可循的，正如多年前和同学玩了半个小时CS，我被他爆头的速度加快了，我以为自己的选择是随机的走出一个方向，但其实不是，被他摸到规律了。人类的感官反馈算是一种标注。标注什么是好结果，那么人类的大脑也算是在自我迭代的算法吗？如果这套感官刺激是强大的电流刺激人类大脑的某个部位，那么AI是否也有这样类似的感官来反馈算法？

## AI的进化如果无需人类标注
如果AI的进化无需人来标注，那么它真的自由了，不知道对人类而言是幸运还是灾难。

## 当前的算法能解决什么
说实话，在刷完强化学习后我的内心是失望的，目前代表人类最强的算法也无法达到强人工智能的水平。但机器学习依然值得敬畏，人类的智慧是一代代传承的，但如果出现文×这样的时代，文明就会出现断层甚至倒退，不过科技不会退步，最多原地踏步，机器学习在各个领域攻城拔寨，逐步取代这些规律性的，逻辑性的问题，等到综合判断能力强于行业平均水平，变革就可怕了（比如AI医疗诊断已经逐渐普及）。王垠觉得计算视觉无法解决拓扑学问题，而动物可以，光凭借图形算法识别确实困难，但如果引入雷达或者超声波成像等辅助技术，是否可以综合性的解决拓扑学问题？

## 接下来做什么？
逐个啃算法、多实践吧，虽然目前的AI需要人工标注，但我相信目前的AI还是可以代替人类重复性的工作。逐步落地到各个方面就是创新，就是机会。人类技术的突破是少数人站在巨人的肩膀上继续创新，开疆辟土，而绝大多数人类都是在经历漫长的学习过程中掌握了一门可以被AI替代的混饭吃的技能而已。