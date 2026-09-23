import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import BackButton from '@/components/Dashboard/BackButton';

const mockPush = jest.fn();

jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

describe('BackButton', () => {
  it('показывает текст «На главную»', () => {
    render(<BackButton />);
    expect(screen.getByText('На главную')).toBeInTheDocument();
  });

  it('переходит на главную при клике', () => {
    render(<BackButton />);
    fireEvent.click(screen.getByText('На главную'));
    expect(mockPush).toHaveBeenCalledWith('/');
  });
});