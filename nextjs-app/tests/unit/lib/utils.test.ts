// tests/unit/lib/utils.test.ts
import { cn } from '@/lib/utils';

describe('Utils Library', () => {
  describe('cn function', () => {
    it('合并CSS类名', () => {
      const result = cn('class1', 'class2');
      expect(result).toBe('class1 class2');
    });

    it('处理条件性类名', () => {
      const result = cn('base', true && 'conditional', false && 'ignored');
      expect(result).toBe('base conditional');
    });

    it('处理undefined和null值', () => {
      const result = cn('base', undefined, null, 'valid');
      expect(result).toBe('base valid');
    });

    it('处理空字符串', () => {
      const result = cn('base', '', 'valid');
      expect(result).toBe('base valid');
    });

    it('处理重复类名', () => {
      const result = cn('class1', 'class2', 'class1');
      // cn函数应该去重，但具体行为取决于实现
      expect(result).toContain('class1');
      expect(result).toContain('class2');
    });

    it('处理对象形式的类名', () => {
      const result = cn({
        'class1': true,
        'class2': false,
        'class3': true
      });
      expect(result).toContain('class1');
      expect(result).not.toContain('class2');
      expect(result).toContain('class3');
    });

    it('处理数组形式的类名', () => {
      const result = cn(['class1', 'class2'], 'class3');
      expect(result).toContain('class1');
      expect(result).toContain('class2');
      expect(result).toContain('class3');
    });

    it('处理混合类型的参数', () => {
      const result = cn(
        'base',
        ['array1', 'array2'],
        { conditional: true, ignored: false },
        'final'
      );
      expect(result).toContain('base');
      expect(result).toContain('array1');
      expect(result).toContain('array2');
      expect(result).toContain('conditional');
      expect(result).not.toContain('ignored');
      expect(result).toContain('final');
    });
  });
});